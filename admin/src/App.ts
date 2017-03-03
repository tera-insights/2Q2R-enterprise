import { AuthSrvc } from './services/AuthSrvc';
import { StatsSrvc } from './services/StatsSrvc';
import { AppSrvc } from './services/AppSrvc';
import { ServerSrvc } from './services/ServerSrvc';
import { AdminSrvc } from './services/AdminSrvc';
import { UserSrvc } from './services/UserSrvc'
import { MainCtrl } from './controllers/MainCtrl';
import { DashboardCtrl } from './controllers/DashboardCtrl';
import { AdminsCtrl } from './controllers/AdminsCtrl';
import { RegisterCtrl } from './controllers/modals/RegisterCtrl';
import { LoginCtrl } from './controllers/LoginCtrl';
import { AppsCtrl } from './controllers/AppsCtrl';
import { AddAppCtrl } from './controllers/modals/AddAppCtrl';
import { ServersCtrl } from './controllers/ServersCtrl';
import { AddServerCtrl } from './controllers/modals/AddServerCtrl';
import angular = require('angular');
import 'ui-router-extras';

angular.module('2Q2R', [
    'ngAria', 'ngMaterial', 'ngResource', 'ngMessages', 'angular-secure-password',
    'ui.router', 'ngAnimate','ct.ui.router.extras', 'md.data.table', 'ngFileUpload'
])
    .service('AuthSrvc', AuthSrvc)
    .service('StatsSrvc', StatsSrvc)
    .service('AppSrvc', AppSrvc)
    .service('ServerSrvc', ServerSrvc)
    .service('AdminSrvc', AdminSrvc)
    .service('UserSrvc', UserSrvc)
    .controller('MainCtrl', MainCtrl)
    .controller('DashboardCtrl', DashboardCtrl)
    .controller('AdminsCtrl', AdminsCtrl)
    .controller('LoginCtrl', LoginCtrl)
    .controller('RegisterCtrl', RegisterCtrl)
    .controller('AppsCtrl', AppsCtrl)
    .controller('AddAppCtrl', AddAppCtrl)
    .controller('ServersCtrl', ServersCtrl)
    .controller('AddServerCtrl', AddServerCtrl)
    .config((
        $stateProvider: angular.ui.IStateProvider,
        $urlRouterProvider: angular.ui.IUrlRouterProvider
    ) => {
        $urlRouterProvider.otherwise("/login");
        $stateProvider
            // LOGIN
            .state('login', {
                url: "/login",
                template: "<ui-view />",
                controller: "LoginCtrl",
                controllerAs: "ctrl",
                deepStateRedirect: {
                    default: { state: 'login.main' },
                    fn: ($dsr$) => {
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
            // DASHBOARD
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