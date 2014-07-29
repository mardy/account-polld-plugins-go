#ifndef __QTCONTACTS_HPP_
#define __QTCONTACTS_HPP_

#include <QObject>

class Avatar : QObject {
    Q_OBJECT
    public:
        void getThumbnail(char *email);

    public slots:
        void retrieveThumbnail(char *email);
};

#endif
