/// <reference path="../../typings/index.d.ts" />

import * as _ from 'lodash';
import { Servers, IServerItem, IServerResource } from '../services/Servers';
import { Apps, IAppItem, IAppResource } from '../services/Apps';
import { AddServerCtrl } from './modals/AddServerCtrl';

export class ServersCtrl {
    private Server: IServerResource;
    private App: IAppResource;

    private apps: IAppItem[] = [];
    private servers: IServerItem[] = [];
    private selectedServers: IServerItem[] = [];
    private selectedAppID: string = null;

    // Aux datascructures to organize Apps and Servers
    private serversByAppID: { [appID: string] : IServerItem[] } = {};

    selectApp(app: IAppItem) {
        this.selectedServers = this.serversByAppID[app.appID]; 
        this.selectedAppID = app.appID;
    }

    newServer() {
        this.$mdDialog.show({
            controller: AddServerCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/AddServer.html',
            clickOutsideToClose: true,
            locals: {
                apps: this.apps
            }
        }).then((server: IServerItem) => {
            server.$save().then((newServer: IServerItem) => {
                this.servers.push(newServer);
            });
        }, () => {
            // TODO: Add nofitifications
        });
    }

    updateServer(server: IServerItem) {

    }

    removeServer(server: IServerItem) {

    }

    // Get all the info from backend again
    refresh() {
        this.apps = this.App.query() ;

        this.servers = this.Server.query( () => {
            this.serversByAppID = _.groupBy(this.servers, 'appID');
        });
    }

    static $inject = [
        '$mdDialog',
        'Apps',
        'Servers'
    ];
    constructor(
        private $mdDialog: ng.material.IDialogService,
        AppsSrvc: Apps,
        ServersSrvc: Servers
    ) {
        this.Server = ServersSrvc.resource;
        this.App = AppsSrvc.resource;
        this.refresh();
    }

}