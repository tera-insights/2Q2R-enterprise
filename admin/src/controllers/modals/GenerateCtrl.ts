import { AppSrvc, IAppItem, IAppResource } from '../../services/AppSrvc';
import { ServerSrvc, IServerItem, IServerResource } from '../../services/ServerSrvc';
import 'angular-resource';
import 'angular-material';

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
        'AppSrvc',
        'ServerSrvc'
    ];
    constructor(
        private $mdDialog: ng.material.IDialogService,
        private AppsSrvc: AppSrvc,
        private ServersSrvc: ServerSrvc,
    ) {
        this.App = AppsSrvc.resource;
        this.Server = ServersSrvc.resource;
    }

}