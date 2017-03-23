import * as _ from 'lodash';
import { ServerSrvc, IServerInfo } from '../services/ServerSrvc';
import { AppSrvc, IAppInfo } from '../services/AppSrvc';
import { DeleteServersCtrl } from './modals/DeleteServersCtrl';

interface IServerExtendedInfo extends IServerInfo {
    appName?: string;
    userCount?: number;
}

export class ServersCtrl {
    private servers: IServerExtendedInfo[];
    private apps: IAppInfo[];

    private appsByID: { [appID: string]: IAppInfo } = {};
    private selectedServers: IServerExtendedInfo[] = [];
    private filteredServers: IServerExtendedInfo[] = [];

    // These correspond to the different filter types indicated by the drop-down.
    private filterProperty: 'serverName' | 'appName' | 'userCount';
    private filters: { [att: string]: any } = {
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
    }

    // These correspond to the actual filter values the user inputs.
    private maxUsers: number = 0;
    private filterString: string = "";
    private caseSensitive: boolean = false;
    private filterRange: { min: number, max: number } = { min: 0, max: 0 };

    private orderBy: string;
    private state: 'searchClosed' | 'searchOpen' = 'searchClosed';
    private serverFilter: string;
    private appFilter: string;
    private pagination = {
        pageLimit: 11,
        page: 1
    }

    // Aux datascructures to organize Apps and Servers
    private serversByAppID: { [appID: string]: IServerInfo[] } = {};

    updateServer(server: IServerInfo) {

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
        console.log(this.filterProperty)
        var filterFct: (IServerExtendedItem) => boolean;
        var filterObj = this.filters[this.filterProperty];
        if (filterObj) {
            switch (filterObj.type) {
                case "string":
                    filterFct = (server: IServerExtendedInfo): boolean => {
                        var serverProperty = server[this.filterProperty] as string;
                        serverProperty = serverProperty ? serverProperty : ""; // TODO: every server should have every required property
                        return (this.caseSensitive ? serverProperty : 
                            serverProperty.toLowerCase())
                            .indexOf(this.caseSensitive ? this.filterString : 
                                this.filterString.toLowerCase()) != -1;
                    };
                    break;
                case "number":
                    filterFct = (server: IServerExtendedInfo): boolean => {
                        var val = server[this.filterString];
                        return (val >= this.filterRange.min) &&
                            (val <= this.filterRange.max);
                    };
                    break;

                default:  // Cannot apply this, return true
                    filterFct = (server: IServerExtendedInfo): boolean => {
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
        this.$q.all([
            this.ServerSrvc.query().then( (servers) => {
                this.servers = servers;
            }),
            this.AppSrvc.query().then( (apps) => {
                this.apps = apps;
            })
        ]).then(() => {
            this.appsByID = _.keyBy(this.apps, 'appID');
            this.servers.forEach((server: IServerExtendedInfo) => {
                if (server.userCount && server.userCount > this.maxUsers) {
                    this.maxUsers = server.userCount;
                    this.filterRange.max = this.maxUsers;
                }

                var app = this.appsByID[server.appID];
                if (app) {
                    server.appName = app.appName;
                }
                server.userCount = 0; // TODO: remove, we need to actually query user count for each server correctly
            });
            this.applyFilters();
        })
    }

    static $inject = [
        '$q',
        '$mdDialog',
        '$mdToast',
        'AppSrvc',
        'ServerSrvc'
    ];
    
    constructor(
        private $q: ng.IQService,
        private $mdDialog: ng.material.IDialogService,
        private $mdToast: ng.material.IToastService,
        private AppSrvc: AppSrvc,
        private ServerSrvc: ServerSrvc
    ) {
        this.refresh();
    }

}