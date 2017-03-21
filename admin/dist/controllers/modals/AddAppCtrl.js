"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var AddAppCtrl = (function () {
    function AddAppCtrl($mdDialog, AppsSrvc) {
        this.$mdDialog = $mdDialog;
        var App = AppsSrvc.resource;
        this.app = new App({
            appName: ""
        });
    }
    AddAppCtrl.prototype.accept = function () {
        this.$mdDialog.hide(this.app);
    };
    AddAppCtrl.prototype.cancel = function () {
        this.$mdDialog.cancel();
    };
    return AddAppCtrl;
}());
AddAppCtrl.$inject = [
    '$mdDialog',
    'AppSrvc'
];
exports.AddAppCtrl = AddAppCtrl;

//# sourceMappingURL=AddAppCtrl.js.map
