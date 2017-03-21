"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var AddAdminCtrl_1 = require("./modals/AddAdminCtrl");
var AdminsCtrl = (function () {
    function AdminsCtrl(AdminSrvc, $mdDialog, $mdToast) {
        this.AdminSrvc = AdminSrvc;
        this.$mdDialog = $mdDialog;
        this.$mdToast = $mdToast;
        this.filterOpen = false;
        this.filters = [
            "name",
            "status",
            "email",
            "role"
        ];
        this.pagination = {
            pageLimit: 11,
            page: 1
        };
        this.refresh();
        this.applyFilters();
    }
    AdminsCtrl.prototype.applyFilters = function () {
        var _this = this;
        if (this.selectedFilter && this.filterValue) {
            this.filteredAdmins = this.admins.filter(function (admin) {
                return admin[_this.selectedFilter].includes(_this.filterValue);
            });
        }
        else {
            this.filteredAdmins = this.admins;
        }
    };
    AdminsCtrl.prototype.refresh = function () {
        var _this = this;
        this.AdminSrvc.query().then(function (admins) {
            _this.admins = admins;
        });
    };
    AdminsCtrl.prototype.deleteSelected = function () {
    };
    AdminsCtrl.prototype.referAdmin = function () {
        var _this = this;
        this.$mdDialog.show({
            controller: AddAdminCtrl_1.AddAdminCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/AddAdmin.html',
            clickOutsideToClose: true
        }).then(function (error) {
            if (!error) {
                _this.$mdToast.showSimple('Admin successfully registered!');
            }
            else {
                _this.$mdToast.showSimple(error.message);
            }
        }, function () {
        });
    };
    return AdminsCtrl;
}());
AdminsCtrl.$inject = [
    'AdminSrvc',
    '$mdDialog',
    '$mdToast'
];
exports.AdminsCtrl = AdminsCtrl;

//# sourceMappingURL=AdminsCtrl.js.map
