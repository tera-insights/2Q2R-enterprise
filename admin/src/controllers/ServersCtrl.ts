/// <reference path="../../typings/index.d.ts" />

import * as _ from 'lodash';
import { Servers, IServerItem, IServerResource } from '../services/Servers';
import { Apps, IAppItem, IAppResource } from '../services/Apps';
import { DeleteServersCtrl } from './modals/DeleteServersCtrl';

export class ServersCtrl {
    private Server: IServerResource;
    private servers: any[] = [{
        serverName: 'Fake Server #1',
        serverID: 'iuef398vn893n8nf',
        appID: 'owiefwinfewufnwf8Ha'
    }, {
        serverName: 'Fake Server #2',
        serverID: '2oifn2oifon2f2==',
        appID: 'DHWjwndjwndwjn5='
    }, {
        serverName: 'Fake Server #3',
        serverID: '29f8j2Hu2822if3j2=',
        appID: 'ofwenufiwuUN28822'
    }, {
        serverName: 'Fake Server #4',
        serverID: '28dj2d82duewndweu',
        appID: 'eoifnefn828D9ejjk=='
    }];
    private selectedServers: IServerItem[] = [];
    private orderBy: string;
    private state: 'searchClosed' | 'searchOpen' = 'searchClosed';

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