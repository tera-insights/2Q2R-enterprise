"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var RegisterCtrl_1 = require("./modals/RegisterCtrl");
var LoginCtrl = (function () {
    function LoginCtrl(AuthSrvc, $mdDialog, $mdToast) {
        this.AuthSrvc = AuthSrvc;
        this.$mdDialog = $mdDialog;
        this.$mdToast = $mdToast;
        this.password = new Uint8Array(50);
    }
    LoginCtrl.prototype.uploadSigningKey = function (file) {
        var _this = this;
        var fileReader = new FileReader();
        fileReader.onload = function (event) {
            _this.signingKey = JSON.parse(event.target.result);
        };
        fileReader.readAsText(file);
    };
    LoginCtrl.prototype.login = function (userID) {
        this.AuthSrvc.prepareFirstFactor(this.signingKey, userID, this.password);
    };
    LoginCtrl.prototype.register = function () {
        var _this = this;
        this.$mdDialog.show({
            controller: RegisterCtrl_1.RegisterCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/Register.html',
            clickOutsideToClose: true
        }).then(function () {
            _this.$mdToast.showSimple('Registration request saved. Email it to your superadmin to get approved.');
        }, function () {
            _this.$mdToast.showSimple('Registration canceled.');
        });
    };
    return LoginCtrl;
}());
LoginCtrl.$inject = [
    'AuthSrvc',
    '$mdDialog',
    '$mdToast'
];
exports.LoginCtrl = LoginCtrl;

//# sourceMappingURL=LoginCtrl.js.map
