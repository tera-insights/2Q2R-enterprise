"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var _ = require("lodash");
require("angular-resource");
var SubscriptionManager = (function () {
    function SubscriptionManager() {
        this.typeMap = {
            __all__: {}
        };
        this.id = 0;
    }
    SubscriptionManager.prototype.subscribe = function (types, handler) {
        var _this = this;
        this.id++;
        if (!types) {
            this.typeMap['__all__'][this.id] = handler;
        }
        else if (types.length > 0) {
            types.forEach(function (type) {
                if (!_this.typeMap[type])
                    _this.typeMap[type] = {};
                var grp = _this.typeMap[type];
                grp[_this.id] = handler;
            });
        }
        else {
            throw new Error("Subscriber must subscribe to at least one message type or provide 'null' to suscribe to all types.");
        }
        return this.id;
    };
    SubscriptionManager.prototype.unsubscribe = function (ids) {
        for (var grp in this.typeMap) {
            for (var id in ids) {
                delete this.typeMap[grp][id];
            }
        }
    };
    SubscriptionManager.prototype.process = function (msgs) {
        var msgGrps = _.groupBy(msgs, '__type__');
        for (var msgType in msgGrps) {
            var grp = this.typeMap[msgType];
            for (var id in grp) {
                grp[id](msgGrps[msgType]);
            }
            for (var id in this.typeMap['__all__']) {
                this.typeMap['__all__'][id](msgGrps[msgType]);
            }
        }
    };
    return SubscriptionManager;
}());
var StatsSrvc = (function () {
    function StatsSrvc($resource) {
        var _this = this;
        this.$resource = $resource;
        this.subMngr = new SubscriptionManager();
        this.resource = this.$resource("/admin/stats/recent");
        var ws = new WebSocket('ws://' + window.location.host + '/admin/stats/listen');
        ws.onmessage = function (msg) {
            console.log(msg);
            _this.subMngr.process(msg.data);
        };
    }
    StatsSrvc.prototype.subscribe = function (types, handler) {
        handler(this.resource.query());
        return this.subMngr.subscribe(types, handler);
    };
    StatsSrvc.prototype.unsubscribe = function (ids) {
        this.subMngr.unsubscribe(ids);
    };
    return StatsSrvc;
}());
StatsSrvc.$inject = [
    '$resource'
];
exports.StatsSrvc = StatsSrvc;

//# sourceMappingURL=StatsSrvc.js.map
