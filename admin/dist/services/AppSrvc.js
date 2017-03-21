"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
require("angular-resource");
var AppSrvc = (function () {
    function AppSrvc($resource, $q) {
        this.$resource = $resource;
        this.$q = $q;
        this.resource = this.$resource('', {}, {
            'query': { method: 'GET', url: '/admin/app', isArray: true },
            'create': { method: 'POST', url: '/admin/app' },
            'update': { method: 'POST', url: '/admin/app/:id' },
            'delete': { method: 'DELETE', url: '/admin/app/:id' }
        });
    }
    AppSrvc.prototype.query = function () {
        return this.resource.query().$promise;
    };
    AppSrvc.prototype.create = function (req) {
        return this.resource.create({}, req).$promise;
    };
    AppSrvc.prototype.update = function (req) {
        return this.resource.update({ id: '@appID' }, req).$promise;
    };
    AppSrvc.prototype.delete = function (appID) {
        return this.resource.delete({ id: appID }).$promise;
    };
    return AppSrvc;
}());
AppSrvc.$inject = [
    '$resource',
    '$q'
];
exports.AppSrvc = AppSrvc;

//# sourceMappingURL=AppSrvc.js.map
