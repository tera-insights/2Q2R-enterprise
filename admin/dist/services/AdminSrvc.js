"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
require("angular");
require("angular-resource");
var AdminSrvc = (function () {
    function AdminSrvc($resource, $q) {
        this.$resource = $resource;
        this.$q = $q;
        this.resource = this.$resource('', {}, {
            'query': { method: 'GET', url: '/admin/admin', isArray: true },
            'create': { method: 'POST', url: '/admin/new' },
            'roles': { method: 'POST', url: 'admin/admin/roles' },
            'update': { method: 'PUT', url: '/admin/admin/:id' },
            'delete': { method: 'DELETE', url: '/admin/admin/:id' }
        });
    }
    AdminSrvc.prototype.query = function () {
        return this.resource.query().$promise;
    };
    AdminSrvc.prototype.create = function (req) {
        return this.resource.create({}, req).$promise;
    };
    AdminSrvc.prototype.changeRoles = function (req) {
        return this.resource.roles({}, req).$promise;
    };
    AdminSrvc.prototype.update = function (req) {
        return this.resource.update({ id: '@adminID' }, req).$promise;
    };
    AdminSrvc.prototype.delete = function (adminID) {
        return this.resource.delete({ id: adminID }).$promise;
    };
    return AdminSrvc;
}());
AdminSrvc.$inject = [
    '$resource',
    '$q'
];
exports.AdminSrvc = AdminSrvc;

//# sourceMappingURL=AdminSrvc.js.map
