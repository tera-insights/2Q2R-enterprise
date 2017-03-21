"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var AddServerCtrl = (function () {
    function AddServerCtrl($mdDialog, ServersSrvc) {
        this.$mdDialog = $mdDialog;
        this.ServersSrvc = ServersSrvc;
        this.availablePermissions = [
            'Register',
            'Authenticate',
            'Analytics'
        ];
        this.selectedPermissions = [];
        var Server = ServersSrvc.resource;
        this.server = new Server({
            serverName: "",
            appID: "",
            baseURL: "",
            keyType: "P-256",
            publicKey: "",
            permissions: ""
        });
    }
    AddServerCtrl.prototype.accept = function () {
        this.server.permissions = this.selectedPermissions.join(',');
        this.$mdDialog.hide(this.server);
    };
    AddServerCtrl.prototype.cancel = function () {
        this.$mdDialog.cancel();
    };
    return AddServerCtrl;
}());
AddServerCtrl.$inject = [
    '$mdDialog',
    'ServerSrvc'
];
exports.AddServerCtrl = AddServerCtrl;

//# sourceMappingURL=AddServerCtrl.js.map
