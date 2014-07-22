#include "_cgo_export.h"

AccountWatcher *watch_for_services(void *array_of_service_names, int length) {
    /* Transfer service names to hash table */
    GoString *service_names = (GoString *)array_of_service_names;
    GHashTable *services_to_watch = g_hash_table_new_full(
        g_str_hash, g_str_equal, g_free, NULL);
    int i;
    for (i = 0; i < length; i++) {
        g_hash_table_insert(services_to_watch, g_strdup(service_names[i].p), NULL);
    }

    if (FALSE) {
        /* The Go callback doesn't quite match the
         * AccountEnabledCallback function prototype, so we cast the
         * argument in the account_watcher_new() call below.
         *
         * This is just a check to see that the function still has the
         * prototype we expect.
         */
        void (*unused)(void *watcher,
                       unsigned int account_id, char *service_name, int enabled,
                       char *client_id, char *client_secret,
                       char *access_token, char *token_secret,
                       void *user_data) = authCallback;
    }

    AccountWatcher *watcher = account_watcher_new(
        services_to_watch, (AccountEnabledCallback)authCallback, NULL);
    g_hash_table_unref(services_to_watch);
    return watcher;
}
