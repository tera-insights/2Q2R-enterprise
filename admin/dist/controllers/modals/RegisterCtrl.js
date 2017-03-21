"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var p256_auth_1 = require("p256-auth");
require("angular-material");
var FileSaver = require("file-saver");
var RegisterCtrl = (function () {
    function RegisterCtrl($mdDialog) {
        this.$mdDialog = $mdDialog;
        this.availablePermissions = [
            'Register',
            'Authenticate',
            'Analytics'
        ];
        this.registration = {
            name: '',
            email: '',
            adminFor: '',
            iv: undefined,
            salt: undefined,
            publicKey: undefined
        };
        this.password = new Uint8Array(50);
        console.log(this.password[0]);
    }
    RegisterCtrl.prototype.accept = function () {
        var _this = this;
        var authenticator = p256_auth_1.createAuthenticator();
        authenticator.generateKeyPair().then(function () {
            Promise.all([authenticator.exportKey(_this.password), authenticator.getPublic()]).then(function (_a) {
                var extKey = _a[0], pubKey = _a[1];
                _this.registration.iv = extKey.iv;
                _this.registration.salt = extKey.salt;
                _this.registration.publicKey = pubKey;
                var keyFile = new Blob([JSON.stringify(extKey, null, 2)], { type: 'text/json;charset=utf-8' });
                var regFile = new Blob([JSON.stringify(_this.registration, null, 2)], { type: 'text/json;charset=utf-8' });
                FileSaver.saveAs(keyFile, 'Key.1fa');
                FileSaver.saveAs(regFile, _this.registration.name.replace(' ', '_') + '.arr');
                _this.$mdDialog.hide();
            });
        });
    };
    RegisterCtrl.prototype.cancel = function () {
        this.$mdDialog.cancel();
    };
    return RegisterCtrl;
}());
RegisterCtrl.$inject = [
    '$mdDialog'
];
exports.RegisterCtrl = RegisterCtrl;

//# sourceMappingURL=RegisterCtrl.js.map
