"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
require("angular-resource");
var p256_auth_1 = require("p256-auth");
var AuthSrvc = (function () {
    function AuthSrvc($resource, $q) {
        this.$resource = $resource;
        this.$q = $q;
        this.resource = this.$resource('', { headers: this.authHeaders }, {
            public: { method: 'GET', url: '/admin/public' },
            nonce: { method: 'GET', url: '/admin/nonce/:id' }
        });
        this.ephemeralKeyManager = p256_auth_1.createAuthenticator();
        this.signingKeyManager = p256_auth_1.createSigner();
    }
    AuthSrvc.prototype.prepareFirstFactor = function (key, id, password) {
        var _this = this;
        var deferred = this.$q.defer();
        this.adminID = id;
        this.$q.all([
            this.getServerPublic(),
            this.getNonce()
        ]).then(function (_a) {
            var serverPublic = _a[0], nonce = _a[1];
            Promise.all([
                _this.ephemeralKeyManager.generateKeyPair(),
                _this.ephemeralKeyManager.importServerKey(serverPublic.public),
                _this.signingKeyManager.importKey(key, password)
            ]).then(function () {
                _this.ephemeralKeyManager.getPublic().then(function (ephemeralPublic) {
                    _this.signingKeyManager.sign(ephemeralPublic, 'base64URL').then(function (signature) {
                        _this.authHeaders = {
                            'X-Authentication-Type': 'admin-frontend',
                            'X-Public-Key': ephemeralPublic,
                            'X-Public-Signature': signature
                        };
                        deferred.resolve();
                    });
                });
            });
        });
        return deferred.promise;
    };
    AuthSrvc.prototype.getAuthHeaders = function (message) {
        var _this = this;
        var deferred = this.$q.defer();
        var msgString = JSON.stringify(message);
        this.getNonce().then(function (nonce) {
            _this.ephemeralKeyManager.computeHMAC(msgString).then(function (hmac) {
                _this.authHeaders['X-Nonce'] = nonce.nonce;
                _this.authHeaders['X-authentication'] = _this.adminID + ':' + hmac;
                deferred.resolve(_this.authHeaders);
            });
        });
        return deferred.promise;
    };
    AuthSrvc.prototype.getServerPublic = function () {
        return this.resource.public().$promise;
    };
    AuthSrvc.prototype.getNonce = function () {
        return this.resource.nonce({ id: this.adminID }).$promise;
    };
    return AuthSrvc;
}());
AuthSrvc.$inject = [
    '$resource',
    '$q'
];
exports.AuthSrvc = AuthSrvc;

//# sourceMappingURL=AuthSrvc.js.map
