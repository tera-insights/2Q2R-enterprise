"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var _ = require("lodash");
var DeleteServersCtrl_1 = require("./modals/DeleteServersCtrl");
require("angular-resource");
require("angular-material");
var ServersCtrl = (function () {
    function ServersCtrl($q, $mdDialog, $mdToast, AppSrvc, ServerSrvc) {
        this.$q = $q;
        this.$mdDialog = $mdDialog;
        this.$mdToast = $mdToast;
        this.AppSrvc = AppSrvc;
        this.ServerSrvc = ServerSrvc;
        this.appsByID = {};
        this.selectedServers = [];
        this.filteredServers = [];
        this.filters = {
            "serverName": {
                type: "string",
                name: "Server Name"
            },
            "appName": {
                type: "string",
                name: "Application Name"
            },
            "userCount": {
                type: "number",
                name: "User Count"
            }
        };
        this.maxUsers = 0;
        this.filterString = "";
        this.caseSensitive = false;
        this.filterRange = { min: 0, max: 0 };
        this.state = 'searchClosed';
        this.pagination = {
            pageLimit: 11,
            page: 1
        };
        this.serversByAppID = {};
        this.App = AppSrvc.resource;
        this.Server = ServerSrvc.resource;
        this.refresh();
    }
    ServersCtrl.prototype.updateServer = function (server) {
    };
    ServersCtrl.prototype.deleteSelected = function () {
        var _this = this;
        this.$mdDialog.show({
            controller: DeleteServersCtrl_1.DeleteServersCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/DeleteServers.html',
            clickOutsideToClose: true
        }).then(function () {
            _this.$mdToast.showSimple('Deleted ' + _this.selectedServers.length + ' ' + (_this.selectedServers.length > 1 ? 'servers' : 'server') + '!');
        }, function () {
        });
    };
    ServersCtrl.prototype.applyFilters = function () {
        var _this = this;
        console.log(this.filterProperty);
        var filterFct;
        var filterObj = this.filters[this.filterProperty];
        if (filterObj) {
            switch (filterObj.type) {
                case "string":
                    filterFct = function (server) {
                        var serverProperty = server[_this.filterProperty];
                        serverProperty = serverProperty ? serverProperty : "";
                        return (_this.caseSensitive ? serverProperty : serverProperty.toLowerCase()).includes(_this.caseSensitive ? _this.filterString : _this.filterString.toLowerCase());
                    };
                    break;
                case "number":
                    filterFct = function (server) {
                        var val = server[_this.filterString];
                        return (val >= _this.filterRange.min) &&
                            (val <= _this.filterRange.max);
                    };
                    break;
                default:
                    filterFct = function (server) {
                        return true;
                    };
            }
            this.filteredServers = this.servers.filter(filterFct);
        }
        else {
            this.filteredServers = this.servers;
        }
    };
    ServersCtrl.prototype.refresh = function () {
        var _this = this;
        console.log("Refreshed servers!");
        this.servers = this.Server.query();
        this.apps = this.App.query();
        this.$q.all([
            this.servers.$promise,
            this.apps.$promise
        ]).then(function () {
            _this.appsByID = _.keyBy(_this.apps, 'appID');
            _this.servers.forEach(function (server) {
                if (server.userCount && server.userCount > _this.maxUsers) {
                    _this.maxUsers = server.userCount;
                    _this.filterRange.max = _this.maxUsers;
                }
                var app = _this.appsByID[server.appID];
                if (app) {
                    server.appName = app.appName;
                }
                server.userCount = 0;
            });
            _this.applyFilters();
        });
    };
    return ServersCtrl;
}());
ServersCtrl.$inject = [
    '$q',
    '$mdDialog',
    '$mdToast',
    'AppSrvc',
    'ServerSrvc'
];
exports.ServersCtrl = ServersCtrl;

//# sourceMappingURL=ServersCtrl.js.map
