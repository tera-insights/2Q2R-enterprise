/// <reference path="../../../typings/index.d.ts" />
/// <reference path="../../definitions/prob.d.ts" />


import { Apps, IAppItem, IAppResource } from '../../services/Apps';
import { Servers, IServerItem, IServerResource } from '../../services/Servers';

export class GenerateCtrl {
    private App: IAppResource;
    private Server: IServerResource;

    private appPrefix: string;
    private serverPrefix: string;
    private numApps = 100;
    private numServers = 1000;

    /**
     * Accept function. Closes modal
     */
    accept() {
        let apps: IAppItem[] = this.App.query();
        
        for (var i = 0; i < this.numServers; i++) {
            let app = apps[Math.floor(Math.random() * apps.length)];
            var server = new this.Server({
                serverName: this.serverPrefix + " #" + (i + 1),
                appID: app.appID,
                baseURL: "",
                keyType: "",
                publicKey: "",
                permissions: ""
            });
            server.$save();
        }

        this.$mdDialog.hide();
    }

    /**
     * Cancel all the actions
     */
    cancel() {
        this.$mdDialog.cancel();
    }

    static $inject = [
        '$mdDialog',
        'Apps',
        'Servers'
    ];
    constructor(
        private $mdDialog: ng.material.IDialogService,
        AppsSrvc: Apps,
        ServersSrvc: Servers,
    ) {
        this.App = AppsSrvc.resource;
        this.Server = ServersSrvc.resource;
    }

}