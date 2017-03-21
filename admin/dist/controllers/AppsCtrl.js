"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var AddAppCtrl_1 = require("../controllers/modals/AddAppCtrl");
require("angular-resource");
require("angular-material");
var AppsCtrl = (function () {
    function AppsCtrl($mdDialog, AppSrvc) {
        this.$mdDialog = $mdDialog;
        this.AppSrvc = AppSrvc;
        this.apps = [];
        this.selected = [];
        this.options = {
            rowSelect: true,
            autoSelect: true,
            multiSelect: true
        };
        this.tableQuery = {
            limit: 14,
            page: 1,
            order: "name"
        };
        this.App = AppSrvc.resource;
        this.apps = this.App.query();
    }
    AppsCtrl.prototype.newApp = function () {
        var _this = this;
        this.$mdDialog.show({
            controller: AddAppCtrl_1.AddAppCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/AddApp.html',
            clickOutsideToClose: true
        }).then(function (app) {
            app.$save().then(function (newApp) {
                _this.apps.push(newApp);
            });
        }, function () {
        });
    };
    AppsCtrl.prototype.updateApp = function (app) {
        app.$update();
    };
    AppsCtrl.prototype.removeApp = function (app) {
        var $index = -1;
        this.apps.forEach(function (t, i, a) {
            if (t.appID == app.appID)
                $index = i;
        });
        if ($index >= 0) {
            app.$delete();
            this.apps.splice($index, 1);
        }
    };
    return AppsCtrl;
}());
AppsCtrl.$inject = [
    '$mdDialog',
    'AppSrvc'
];
exports.AppsCtrl = AppsCtrl;

//# sourceMappingURL=AppsCtrl.js.map
