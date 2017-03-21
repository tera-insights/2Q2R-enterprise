"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var GenerateCtrl_1 = require("./modals/GenerateCtrl");
var MainCtrl = (function () {
    function MainCtrl($mdSidenav, $state, $mdDialog, AuthSrvc) {
        this.$mdSidenav = $mdSidenav;
        this.$state = $state;
        this.$mdDialog = $mdDialog;
        this.AuthSrvc = AuthSrvc;
        this.sName = "";
        this.menuGroups = [
            [
                {
                    state: "main.dashboard",
                    name: "Dashboard",
                    icon: "mdi mdi-view-dashboard"
                },
                {
                    state: "main.admin",
                    name: "Administrators",
                    icon: "mdi mdi-certificate"
                },
                {
                    state: "main.apps",
                    name: "Applications",
                    icon: "mdi mdi-cloud"
                },
                {
                    state: "main.servers",
                    name: "Servers",
                    icon: "mdi mdi-server"
                },
                {
                    state: "main.users",
                    name: "Users",
                    icon: "mdi mdi-account"
                },
                {
                    state: "main.2FA",
                    name: "2FA Devices",
                    icon: "mdi mdi-cellphone-link"
                },
                {
                    state: "main.reports",
                    name: "Reports",
                    icon: "mdi mdi-clipboard-text"
                },
                {
                    state: "main.settings",
                    name: "Settings",
                    icon: "mdi mdi-settings"
                },
                {
                    state: "login",
                    name: "Logout",
                    icon: "mdi mdi-logout-variant"
                }
            ]
        ];
        this.toggleLeft();
        this.select("main.dashboard", "Dashboard");
    }
    MainCtrl.prototype.select = function (route, name, menu) {
        this.sName = name;
        this.$state.go(route);
        this.activeMenu = menu;
        this.$mdSidenav('left').toggle();
    };
    MainCtrl.prototype.toggleLeft = function () {
        this.$mdSidenav('left').toggle();
    };
    MainCtrl.prototype.generate = function () {
        this.$mdDialog.show({
            controller: GenerateCtrl_1.GenerateCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/Generate.html',
            clickOutsideToClose: true
        }).then(function () {
        }, function () {
        });
    };
    return MainCtrl;
}());
MainCtrl.$inject = [
    '$mdSidenav',
    '$state',
    '$mdDialog',
    'AuthSrvc'
];
exports.MainCtrl = MainCtrl;

//# sourceMappingURL=MainCtrl.js.map
