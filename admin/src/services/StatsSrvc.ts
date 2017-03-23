import * as _ from 'lodash';

export type MessageHandler = (msgs: any[]) => void;

export interface Message extends ng.resource.IResource<Message> {
    __type__: string
}

type HandlerMap = { [id: number]: MessageHandler };
type TypeMap = { [type: string]: HandlerMap };

class SubscriptionManager {
    // main map from type to handlers
    private typeMap: TypeMap = {
        __all__: {}
    };
    // request IDs to handler unsubscribe
    private id: number = 0;

    // if @types == null, subscribe to all types
    subscribe(types: string[], handler: MessageHandler) {
        // next id
        this.id++;

        if (!types) {
            this.typeMap['__all__'][this.id] = handler;
        } else if (types.length > 0) {
            types.forEach((type) => {
                if (!this.typeMap[type])
                    this.typeMap[type] = {};
                var grp = this.typeMap[type];

                grp[this.id] = handler;
            });
        } else {
            throw new Error("Subscriber must subscribe to at least one message type or provide 'null' to suscribe to all types.");
        }

        return this.id;
    }

    unsubscribe(ids: Array<number>) {
        for (var grp in this.typeMap) {
            for (let id in ids) {
                delete this.typeMap[grp][id];
            }
        }
    }

    process(msgs: Message[]) {
        let msgGrps = _.groupBy(msgs, '__type__');

        for (let msgType in msgGrps) {
            let grp: HandlerMap = this.typeMap[msgType];

            for (let id in grp) {
                grp[id](msgGrps[msgType]);
            }

            for (let id in this.typeMap['__all__']) {
                this.typeMap['__all__'][id](msgGrps[msgType]);
            }
        }
    }

}

export class StatsSrvc {
    private subMngr: SubscriptionManager = new SubscriptionManager();

    public resource = this.$resource("/admin/stats/recent");

    subscribe(types: string[], handler: MessageHandler) {
        handler(this.resource.query()); // TODO: only give the recent messages for particular types
        return this.subMngr.subscribe(types, handler);
    }

    unsubscribe(ids: number[]) {
        this.subMngr.unsubscribe(ids);
    }

    static $inject = [
        '$resource'
    ];

    constructor(
        private $resource: ng.resource.IResourceService
    ) {
        var ws = new WebSocket('ws://' + window.location.host + '/admin/stats/listen');
        ws.onmessage = (msg) => {
            console.log(msg);
            this.subMngr.process(msg.data);
        };
    }

}