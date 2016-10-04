/// <reference path="../typings/index.d.ts" />
/// <reference path="controllers/MainCtrl.ts" />
/// <reference path="controllers/AdminsCtrl.ts" />

module admin {
    var admin = angular.module('2Q2R', [
        'ngAria', 'ngMaterial', 'ngResource',
        'ui.router', 'ct.ui.router.extras'
    ])
        .controller('MainCtrl', MainCtrl)
        .controller('AdminsCtrl', AdminsCtrl)
        .config((
            $stateProvider: angular.ui.IStateProvider,
            $urlRouterProvider: angular.ui.IUrlRouterProvider
        ) => {
            $urlRouterProvider.otherwise("/main");
            $stateProvider
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
                .state('main', {
                    url: "/main",
                    templateUrl: "views/main.html",
                    controller: "MainCtrl",
                    controllerAs: "ctrl"
                })
                .state('main.admin',{
                    url:"/main/admin",
                    templateUrl: "views/admins.html",
                    controller: "AdminsCtrl",
                    controllerAs: "ctrl2"
                })
                ;

        })
        ;
}