/// <reference path="../../typings/index.d.ts" />

import * as _ from 'lodash';
import { Servers, IServerItem, IServerResource } from '../services/Servers';
import { Apps, IAppItem, IAppResource } from '../services/Apps';
import { DeleteServersCtrl } from './modals/DeleteServersCtrl';

interface IServerExtendedItem extends IServerItem {
    appName: string;
}

type filterAttType = 'serverName' | 'appName' | 'userCnt';


export class ServersCtrl {
    private Server: IServerResource;
    private App: IAppResource;
    private servers: IServerExtendedItem[];
    private apps: IAppItem[];

    private appsByID: { [appID: string]: IAppItem } = {};

    private selectedServers: IServerExtendedItem[] = [];
    private filteredServers: IServerExtendedItem[] = [];

    private filterAtt: filterAttType;
    private filters: { [att: string]: any } = {
        "serverName": {
            type: "string",
            name: "Server Name"
        },
        "appName": {
            type: "string",
            name: "Application Name"
        },
        "userCnt": {
            type: "number",
            name: "User Count"
        }
    }
    private filterString: string = "";
    private filterRange: { min: number, max: number };

    private orderBy: string;
    private state: 'searchClosed' | 'searchOpen' = 'searchClosed';
    private serverFilter: string;
    private appFilter: string;
    private pagination = {
        pageLimit: 11,
        page: 1
    }

    // Aux datascructures to organize Apps and Servers
    private serversByAppID: { [appID: string]: IServerItem[] } = {};

    updateServer(server: IServerItem) {

    }

    deleteSelected() {
        this.$mdDialog.show({
            controller: DeleteServersCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/DeleteServers.html',
            clickOutsideToClose: true
        }).then(() => {
            // TODO: delete selected apps
            this.$mdToast.showSimple('Deleted ' + this.selectedServers.length + ' ' + (this.selectedServers.length > 1 ? 'servers' : 'server') + '!');
        }, () => {
            // user canceled, do nothing
        });
    }

    applyFilters() {
        var filterFct: (IServerExtendedItem) => boolean;
        var filterObj = this.filters[this.filterAtt];
        if (filterObj) {
            switch (filterObj.type) {
                case "string":
                    filterFct = (server: IServerExtendedItem): boolean => {
                        return server[this.filterAtt].includes(this.filterString);
                    };
                    break;
                case "number":
                    filterFct = (server: IServerExtendedItem): boolean => {
                        let val = server[this.filterAtt];
                        return (val >= this.filterRange.min) &&
                            (val <= this.filterRange.max);
                    };
                    break;

                default:  // Cannot apply this, return true
                    filterFct = (server: IServerExtendedItem): boolean => {
                        return true;
                    };
            }
            this.filteredServers = this.servers.filter(filterFct);
        } else {
            this.filteredServers = this.servers;
        }
    }

    // Get all the info from backend again
    refresh() {
        console.log("Refreshed servers!");
        this.servers = <IServerExtendedItem[]>this.Server.query();
        this.apps = this.App.query();

        this.$q.all([
            this.servers.$promise,
            this.apps.$promise
        ]).then(() => {
            this.appsByID = _.keyBy(this.apps, 'appID');
            this.servers.forEach((server: IServerExtendedItem) => {
                var app = this.appsByID[server.appID];
                if (app) {
                    server.appName = app.appName;
                }
            });
            this.applyFilters();
        })
    }

    static $inject = [
        '$q',
        '$mdDialog',
        '$mdToast',
        'Apps',
        'Servers'
    ];
    constructor(
        private $q: ng.IQService,
        private $mdDialog: ng.material.IDialogService,
        private $mdToast: ng.material.IToastService,
        AppsSrvc: Apps,
        ServersSrvc: Servers
    ) {
        this.App = AppsSrvc.resource;
        this.Server = ServersSrvc.resource;
        this.refresh();
    }

}