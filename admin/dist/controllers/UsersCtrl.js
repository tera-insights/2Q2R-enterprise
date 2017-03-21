"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
require("angular-resource");
var UsersCtrl = (function () {
    function UsersCtrl(UsersSrvc) {
        this.users = [];
        this.User = UsersSrvc.resource;
        this.users = this.User.query();
    }
    return UsersCtrl;
}());
UsersCtrl.$inject = [
    'Users'
];
exports.UsersCtrl = UsersCtrl;

//# sourceMappingURL=UsersCtrl.js.map
