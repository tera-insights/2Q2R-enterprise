/// <reference path="../typings/index.d.ts" />
/// <reference path="controllers/MainCtrl.ts" />

module admin {
    var admin = angular.module('2Q2R', [
        'ngAria', 'ngMaterial', 'ngResource', 
        'ui.router', 'ct.ui.router.extras'
    ])
    .controller('MainCtrl', MainCtrl)
    ;
}