"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
require("angular-resource");
var ServerSrvc = (function () {
    function ServerSrvc($resource, $q) {
        this.$resource = $resource;
        this.$q = $q;
        this.resource = this.$resource('', {}, {
            'query': { method: 'GET', url: '/admin/server', isArray: true },
            'create': { method: 'POST', url: '/admin/server', },
            'update': { method: 'PUT', url: '/admin/server/:id' },
            'delete': { method: 'DELETE', url: '/admin/server/:id' }
        });
    }
    ServerSrvc.prototype.query = function () {
        return this.resource.query().$promise;
    };
    ServerSrvc.prototype.create = function (req) {
        return this.resource.create({}, req).$promise;
    };
    ServerSrvc.prototype.update = function (req) {
        return this.resource.update({ id: '@serverID' }, req).$promise;
    };
    ServerSrvc.prototype.delete = function (serverID) {
        return this.resource.delete({ id: serverID }).$promise;
    };
    return ServerSrvc;
}());
ServerSrvc.$inject = [
    '$resource',
    '$q'
];
exports.ServerSrvc = ServerSrvc;

//# sourceMappingURL=ServerSrvc.js.map
