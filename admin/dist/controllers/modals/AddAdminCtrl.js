"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
require("angular-material");
var AddAdminCtrl = (function () {
    function AddAdminCtrl($mdDialog, AdminSrvc) {
        this.$mdDialog = $mdDialog;
        this.AdminSrvc = AdminSrvc;
        this.availablePermissions = [
            'Register',
            'Authenticate',
            'Analytics'
        ];
    }
    AddAdminCtrl.prototype.uploadRegistration = function (file) {
        var _this = this;
        var fileReader = new FileReader();
        fileReader.onload = function (event) {
            _this.registration = JSON.parse(event.target.result);
        };
        fileReader.readAsText(file);
    };
    AddAdminCtrl.prototype.uploadKey = function (file) {
    };
    AddAdminCtrl.prototype.accept = function () {
        var _this = this;
        this.AdminSrvc.create(this.registration).then(function (reply) {
            _this.$mdDialog.hide();
        }).catch(function (error) {
            _this.$mdDialog.hide(error);
        });
    };
    AddAdminCtrl.prototype.cancel = function () {
        this.$mdDialog.cancel();
    };
    return AddAdminCtrl;
}());
AddAdminCtrl.$inject = [
    '$mdDialog',
    'AdminSrvc'
];
exports.AddAdminCtrl = AddAdminCtrl;

//# sourceMappingURL=AddAdminCtrl.js.map
