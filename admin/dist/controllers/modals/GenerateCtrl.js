"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
require("angular-resource");
require("angular-material");
var GenerateCtrl = (function () {
    function GenerateCtrl($mdDialog, AppsSrvc, ServersSrvc) {
        this.$mdDialog = $mdDialog;
        this.AppsSrvc = AppsSrvc;
        this.ServersSrvc = ServersSrvc;
        this.numApps = 100;
        this.numServers = 1000;
        this.App = AppsSrvc.resource;
        this.Server = ServersSrvc.resource;
    }
    GenerateCtrl.prototype.accept = function () {
        var apps = this.App.query();
        for (var i = 0; i < this.numServers; i++) {
            var app = apps[Math.floor(Math.random() * apps.length)];
            var server = new this.Server({
                serverName: this.serverPrefix + " #" + (i + 1),
                appID: app.appID,
                baseURL: "",
                keyType: "",
                publicKey: "",
                permissions: ""
            });
            server.$save();
        }
        this.$mdDialog.hide();
    };
    GenerateCtrl.prototype.cancel = function () {
        this.$mdDialog.cancel();
    };
    return GenerateCtrl;
}());
GenerateCtrl.$inject = [
    '$mdDialog',
    'AppSrvc',
    'ServerSrvc'
];
exports.GenerateCtrl = GenerateCtrl;

//# sourceMappingURL=GenerateCtrl.js.map
