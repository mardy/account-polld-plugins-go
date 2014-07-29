#ifndef __QTCONTACTS_HPP_
#define __QTCONTACTS_HPP_

#include <QObject>
#include <QSignalMapper>

class Avatar : QObject {
    Q_OBJECT
    public:
        explicit Avatar(QObject* parent=0)
            : QObject(parent) {
            _signalMapper = new QSignalMapper(this);
            // connect empty signal
            connect(this, SIGNAL(readyToRetrieve()),
                _signalMapper, SLOT(map()));
            // connect signal that takes string
            connect(_signalMapper, SIGNAL(mapped(QString)),
                this, SLOT(retrieveThumbnail(QString)));
        }

        ~Avatar() {
            _signalMapper->deleteLater();
        }

        void getThumbnail(char *email);

    public slots:
        void emitSignals();
        void retrieveThumbnail(const QString& email);

    signals:
        void readyToRetrieve();

    private:
        QSignalMapper* _signalMapper = nullptr;
};

#endif
