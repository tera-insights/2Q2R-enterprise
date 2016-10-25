/// <reference path="../../typings/index.d.ts" />

import * as _ from 'lodash';
import { Servers, IServerItem, IServerResource } from '../services/Servers';
import { Apps, IAppItem, IAppResource } from '../services/Apps';
import { DeleteServersCtrl } from './modals/DeleteServersCtrl';

export class ServersCtrl {
    private Server: IServerResource;
    private servers: IServerItem[];
    private selectedServers: IServerItem[] = [];
    private orderBy: string;
    private state: 'searchClosed' | 'searchOpen' = 'searchClosed';
    private serverFilter: string;
    private appFilter: string;
    private pagination = {
        pageLimit: 11,
        page: 1
    }

    // Aux datascructures to organize Apps and Servers
    private serversByAppID: { [appID: string] : IServerItem[] } = {};

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

    // Get all the info from backend again
    refresh() {
        console.log("Refreshed servers!");
        this.servers = this.Server.query();
        // this.servers = this.Server.query( () => {
        //     this.serversByAppID = _.groupBy(this.servers, 'appID');
        // });
    }

    static $inject = [
        '$mdDialog',
        '$mdToast',
        'Servers'
    ];
    constructor(
        private $mdDialog: ng.material.IDialogService,
        private $mdToast: ng.material.IToastService,
        ServersSrvc: Servers
    ) {
        this.Server = ServersSrvc.resource;
        this.refresh();
    }

}