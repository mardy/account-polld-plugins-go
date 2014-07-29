#include <QContactManager>
#include <QContactFilter>
#include <QContactEmailAddress>
#include <QContactDetailFilter>
#include <QContactManager>
#include <QContactAvatar>
#include <QCoreApplication>
#include <QScopedPointer>
#include <QTimer>
#include <thread>

#include "qtcontacts.h"
#include "qtcontacts.hpp"
#include "qtcontacts.moc"

#ifdef __cplusplus
extern "C" {
#include "_cgo_export.h"
}
#endif

QTCONTACTS_USE_NAMESPACE

void getAvatar(char *email) {
    QScopedPointer<Avatar> avatar(new Avatar());
    avatar->getThumbnail(email);
}

void Avatar::getThumbnail(char *email) {
    static char empty[1] = {0};
    static char *argv[] = {empty, empty, empty};
    static int argc = 1;
    QCoreApplication mApp(argc, argv);

    // set the map for this object and the passed email
    _signalMapper->setMapping(this, QString(email));

    QTimer::singleShot(0, this, SLOT(emitSignals()));
    mApp.exec();
}


void Avatar::emitSignals() {
    emit readyToRetrieve();
}

void Avatar::retrieveThumbnail(const QString& email) {
    QString avatar;

    QContactManager manager ("galera");
    QContactDetailFilter filter(QContactEmailAddress::match(email));
    QList<QContact> contacts = manager.contacts(filter);
    if(contacts.size() > 0) {
        avatar = contacts[0].detail<QContactAvatar>().imageUrl().path();
    }

    QByteArray byteArray = avatar.toUtf8();
    char* cString = byteArray.data();

    AvatarPath(cString);
}

