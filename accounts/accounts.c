#include "_cgo_export.h"
#include <stdio.h>
#include <string.h>
#include <unistd.h>

#include <glib.h>
#include <libaccounts-glib/accounts-glib.h>
#include <libsignon-glib/signon-glib.h>

typedef struct _AuthContext AuthContext;
struct _AuthContext {
    AgManager *manager;
    char *service_name;
    AgAccountService *account_service;
    AgAuthData *auth_data;
    SignonAuthSession *session;

    GVariant *auth_params;
    GVariant *session_data;
};

static void login_cb(GObject *source, GAsyncResult *result, void *user_data) {
    SignonAuthSession *session = (SignonAuthSession *)source;
    AuthContext *ctx = (AuthContext *)user_data;

    GError *error = NULL;
    ctx->session_data = signon_auth_session_process_finish(session, result, &error);

    g_object_unref(ctx->session);
    ctx->session = NULL;

    if (error != NULL) {
        fprintf(stderr, "Authentication failed: %s\n", error->message);
        g_error_free(error);
        return;
    }

    char *client_id = NULL;
    char *client_secret = NULL;
    char *access_token = NULL;
    g_variant_lookup(ctx->auth_params, "ClientId", "&s", &client_id);
    g_variant_lookup(ctx->auth_params, "ClientSecret", "&s", &client_secret);
    g_variant_lookup(ctx->session_data, "AccessToken", "&s", &access_token);

    authLogin(ctx, client_id, client_secret, access_token);
}

static void login_service(AuthContext *ctx) {
    ctx->auth_data = ag_account_service_get_auth_data(ctx->account_service);

    GError *error = NULL;
    ctx->session = signon_auth_session_new(
        ag_auth_data_get_credentials_id(ctx->auth_data),
        ag_auth_data_get_method(ctx->auth_data), &error);
    if (error != NULL) {
        fprintf(stderr, "Could not set up auth session: %s\n", error->message);
        g_error_free(error);
        return;
    }

    GVariantBuilder builder;
    g_variant_builder_init(&builder, G_VARIANT_TYPE_VARDICT);
    g_variant_builder_add(
        &builder, "{sv}",
        SIGNON_SESSION_DATA_UI_POLICY,
        g_variant_new_int32(SIGNON_POLICY_NO_USER_INTERACTION));

    ctx->auth_params = g_variant_ref_sink(
        ag_auth_data_get_login_parameters(
            ctx->auth_data, g_variant_builder_end(&builder)));

    signon_auth_session_process_async(
        ctx->session,
        ctx->auth_params,
        ag_auth_data_get_mechanism(ctx->auth_data),
        NULL, /* cancellable */
        login_cb, ctx);
}

static void logout_service(AuthContext *ctx) {
    if (ctx->session_data) {
        g_variant_unref(ctx->session_data);
        ctx->session_data = NULL;
    }
    if (ctx->auth_params) {
        g_variant_unref(ctx->auth_params);
        ctx->auth_params = NULL;
    }
    if (ctx->session) {
        signon_auth_session_cancel(ctx->session);
        g_object_unref(ctx->session);
        ctx->session = NULL;
    }
    if (ctx->auth_data) {
        g_object_unref(ctx->auth_data);
        ctx->auth_data = NULL;
    }
    if (ctx->account_service) {
        g_object_unref(ctx->account_service);
        ctx->account_service = NULL;
    }

    authLogin(ctx, NULL, NULL, NULL);
}

static void lookup_account_service(AuthContext *ctx) {
    GList *account_services = ag_manager_get_enabled_account_services(ctx->manager);
    GList *tmp;
    for (tmp = account_services; tmp != NULL; tmp = tmp->next) {
        AgAccountService *acct_svc = AG_ACCOUNT_SERVICE(tmp->data);
        AgService *service = ag_account_service_get_service(acct_svc);
        if (!strcmp(ctx->service_name, ag_service_get_name(service))) {
            ctx->account_service = g_object_ref(acct_svc);
            break;
        }
    }
    g_list_foreach(account_services, (GFunc)g_object_unref, NULL);
    g_list_free(account_services);

    if (ctx->account_service != NULL) {
        login_service(ctx);
    }
}

static void account_enabled_cb(AgManager *manager, guint account_id, void *user_data) {
    AuthContext *ctx = (AuthContext *)user_data;

    printf("enabled_cb account_id=%u\n", account_id);

    if (ctx->account_service != NULL &&
        !ag_account_service_get_enabled(ctx->account_service)) {
        logout_service(ctx);
    }
    lookup_account_service(ctx);
}

static gboolean
setup_context(void *user_data) {
    AuthContext *ctx = (AuthContext *)user_data;

    lookup_account_service(ctx);
    g_signal_connect(ctx->manager, "enabled-event",
                     G_CALLBACK(account_enabled_cb), ctx);
    return FALSE;
}

AuthContext *watch_for_service(const char *service_name) {
    AuthContext *ctx = g_new0(AuthContext, 1);
    ctx->manager = ag_manager_new();
    ctx->service_name = g_strdup(service_name);

    g_idle_add(setup_context, ctx);
    return ctx;
}
