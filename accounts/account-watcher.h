#ifndef ACCOUNT_WATCHER_H
#define ACCOUNT_WATCHER_H

#include <glib.h>

typedef struct _AccountWatcher AccountWatcher;

typedef void (*AccountEnabledCallback)(AccountWatcher *watcher,
                                       unsigned int account_id,
                                       const char *service_name,
                                       int enabled,
                                       const char *client_id,
                                       const char *client_secret,
                                       const char *access_token,
                                       const char *token_secret,
                                       void *user_data);

AccountWatcher *account_watcher_new(GHashTable *services_to_watch,
                                    AccountEnabledCallback callback,
                                    void *user_data);

#endif
