/// <reference path="../../../typings/index.d.ts" />
/// <reference path="../../definitions/prob.d.ts" />


import { Apps, IAppItem, IAppResource } from '../../services/Apps';
import { Servers, IServerItem, IServerResource } from '../../services/Servers';

export class GenerateCtrl {
    private App: IAppResource;
    private Server: IServerResource;

    private appPrefix: string = "App";
    private serverPrefix: string = "Server";
    private numApps = 100;
    private numServers = 1000;
    private zipf = 1;

    /**
     * Accept function. Closes modal
     */
    accept() {
        var dist = Prob.zipf(this.zipf, this.numServers);
        var sNum = 0; // server counter

        for (var i=0; i<this.numApps; i++){
            // create the App first
            var app = new this.App({
                appName: this.appPrefix+"_"+i
            });
            app.$save().then((newApp: IAppItem) => {
                var appID = newApp.appID;
                // Crate the Servers
                var numS: number = dist();
                for (var j=0; j<numS; j++){
                    var server = new this.Server({
                        appID: appID,
                        serverName: this.serverPrefix+" "+this.appPrefix+" "+sNum
                    });
                    server.$save();
                    sNum++; // increment the server number
                }
            });
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