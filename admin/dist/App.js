"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var angular = require("angular");
require("angular-messages");
require("angular-animate");
require("angular-material");
require("angular-aria");
require("angular-ui-router");
require("ui-router-extras");
require("angular-material-data-table");
require("leaflet");
require("chart.js");
require("angular-secure-password");
require("ng-file-upload");
var AuthSrvc_1 = require("./services/AuthSrvc");
var StatsSrvc_1 = require("./services/StatsSrvc");
var AppSrvc_1 = require("./services/AppSrvc");
var ServerSrvc_1 = require("./services/ServerSrvc");
var AdminSrvc_1 = require("./services/AdminSrvc");
var UserSrvc_1 = require("./services/UserSrvc");
var MainCtrl_1 = require("./controllers/MainCtrl");
var DashboardCtrl_1 = require("./controllers/DashboardCtrl");
var AdminsCtrl_1 = require("./controllers/AdminsCtrl");
var RegisterCtrl_1 = require("./controllers/modals/RegisterCtrl");
var LoginCtrl_1 = require("./controllers/LoginCtrl");
var AppsCtrl_1 = require("./controllers/AppsCtrl");
var AddAppCtrl_1 = require("./controllers/modals/AddAppCtrl");
var ServersCtrl_1 = require("./controllers/ServersCtrl");
var AddServerCtrl_1 = require("./controllers/modals/AddServerCtrl");
angular.module('2Q2R', [
    'ngAria', 'ngMaterial', 'ngResource', 'ngMessages', 'angular-secure-password',
    'ui.router', 'ngAnimate', 'ct.ui.router.extras', 'md.data.table', 'ngFileUpload'
])
    .service('AuthSrvc', AuthSrvc_1.AuthSrvc)
    .service('StatsSrvc', StatsSrvc_1.StatsSrvc)
    .service('AppSrvc', AppSrvc_1.AppSrvc)
    .service('ServerSrvc', ServerSrvc_1.ServerSrvc)
    .service('AdminSrvc', AdminSrvc_1.AdminSrvc)
    .service('UserSrvc', UserSrvc_1.UserSrvc)
    .controller('MainCtrl', MainCtrl_1.MainCtrl)
    .controller('DashboardCtrl', DashboardCtrl_1.DashboardCtrl)
    .controller('AdminsCtrl', AdminsCtrl_1.AdminsCtrl)
    .controller('LoginCtrl', LoginCtrl_1.LoginCtrl)
    .controller('RegisterCtrl', RegisterCtrl_1.RegisterCtrl)
    .controller('AppsCtrl', AppsCtrl_1.AppsCtrl)
    .controller('AddAppCtrl', AddAppCtrl_1.AddAppCtrl)
    .controller('ServersCtrl', ServersCtrl_1.ServersCtrl)
    .controller('AddServerCtrl', AddServerCtrl_1.AddServerCtrl)
    .config(function ($stateProvider, $urlRouterProvider) {
    $urlRouterProvider.otherwise("/main");
    $stateProvider
        .state('login', {
        url: "/login",
        template: "<ui-view />",
        controller: "LoginCtrl",
        controllerAs: "ctrl",
        deepStateRedirect: {
            default: { state: 'login.main' },
            fn: function ($dsr$) {
                return { state: 'login.main' };
            }
        }
    })
        .state('login.main', {
        url: '',
        templateUrl: "views/login.html"
    })
        .state('login.2q2r', {
        url: "/2q2r",
        templateUrl: "views/iframe.html"
    })
        .state('main', {
        url: "/main",
        templateUrl: "views/main.html",
        controller: "MainCtrl",
        controllerAs: "ctrl"
    })
        .state('main.dashboard', {
        url: "/dashboard",
        templateUrl: "views/dashboard.html",
        controller: "DashboardCtrl",
        controllerAs: "ctrl2"
    })
        .state('main.admin', {
        url: "/admin",
        templateUrl: "views/admins.html",
        controller: "AdminsCtrl",
        controllerAs: "ctrl2"
    })
        .state('main.apps', {
        url: "/apps",
        templateUrl: "views/apps.html",
        controller: "AppsCtrl",
        controllerAs: "ctrl2"
    })
        .state('main.servers', {
        url: "/servers",
        templateUrl: "views/servers.html",
        controller: "ServersCtrl",
        controllerAs: "ctrl2"
    })
        .state('main.users', {
        url: "/users",
        templateUrl: "views/users.html"
    })
        .state('main.2FA', {
        url: "/2FA",
        templateUrl: "views/2FA.html"
    })
        .state('main.reports', {
        url: "/reports",
        templateUrl: "views/reports.html"
    })
        .state('main.settings', {
        url: "/settings",
        templateUrl: "views/settings.html"
    });
});

//# sourceMappingURL=App.js.map
