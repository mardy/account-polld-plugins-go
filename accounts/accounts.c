#include "_cgo_export.h"
#include "account-watcher.h"

AccountWatcher *watch_for_services(void *array_of_service_names, int length) {
    /* Transfer service names to hash table */
    GoString *service_names = (GoString *)array_of_service_names;
    GHashTable *services_to_watch = g_hash_table_new_full(
        g_str_hash, g_str_equal, g_free, NULL);
    int i;
    for (i = 0; i < length; i++) {
        g_hash_table_insert(services_to_watch, g_strdup(service_names[i].p), NULL);
    }

    AccountWatcher *watcher = account_watcher_new(
        services_to_watch, authCallback, NULL);
    g_hash_table_unref(services_to_watch);
    return watcher;
}
