/// <reference path="../typings/index.d.ts" />

import { Auth } from './services/Auth';
import { Apps } from './services/Apps';
import { Servers } from './services/Servers';
import { MainCtrl } from './controllers/MainCtrl';
import { AdminsCtrl } from './controllers/AdminsCtrl';
import { LoginCtrl } from './controllers/LoginCtrl';
import { AppsCtrl } from './controllers/AppsCtrl';
import { AddAppCtrl } from './controllers/modals/AddAppCtrl';
import { ServersCtrl } from './controllers/ServersCtrl';
import { AddServerCtrl } from './controllers/modals/AddServerCtrl';


angular.module('2Q2R', [
    'ngAria', 'ngMaterial', 'ngResource', 'ngMessages',
    'ui.router', 'ngAnimate','ct.ui.router.extras', 'md.data.table'
])
    .service('Auth', Auth)
    .service('Apps', Apps)
    .service('Servers', Servers)
    .controller('MainCtrl', MainCtrl)
    .controller('AdminsCtrl', AdminsCtrl)
    .controller('LoginCtrl', LoginCtrl)
    .controller('AppsCtrl', AppsCtrl)
    .controller('AddAppCtrl', AddAppCtrl)
    .controller('ServersCtrl', ServersCtrl)
    .controller('AddServerCtrl', AddServerCtrl)
    .config((
        $stateProvider: angular.ui.IStateProvider,
        $urlRouterProvider: angular.ui.IUrlRouterProvider
    ) => {
        $urlRouterProvider.otherwise("/main");
        $stateProvider
            // REGISTRATION
            .state('register', {
                url: "/register",
                template: "<ui-view />",
                controller: 'RegisterCtrl',
                controllerAs: "ctrl",
                deepStateRedirect: {
                    default: { state: 'register.main' }
                }
            })
            .state('register.main', {
                url: '',
                templateUrl: "views/register.html",
            })
            .state('register.2q2r', {
                url: "/2q2r",
                templateUrl: "views/iframe.html"
            })
            .state('register.return', {
                url: "/return",
                templateUrl: "views/register.return.html"
            })
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
                templateUrl: "views/dashboard.html"
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
            })
            ;

    })
    ;