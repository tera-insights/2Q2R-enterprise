"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
require("angular-resource");
var UserSrvc = (function () {
    function UserSrvc($resource, $q) {
        this.$resource = $resource;
        this.$q = $q;
        this.resource = this.$resource('', {}, {});
    }
    return UserSrvc;
}());
UserSrvc.$inject = [
    '$resource',
    '$q'
];
exports.UserSrvc = UserSrvc;

//# sourceMappingURL=UserSrvc.js.map
