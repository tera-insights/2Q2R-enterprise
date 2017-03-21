import { AppSrvc, IAppInfo } from '../../services/AppSrvc';
import { ServerSrvc, IServerInfo } from '../../services/ServerSrvc';
import 'angular-resource';
import 'angular-material';

export class GenerateCtrl {
    private appPrefix: string;
    private serverPrefix: string;
    private numApps = 100;
    private numServers = 1000;
    private apps: IAppInfo[];

    /**
     * Accept function. Closes modal
     */
    accept() {
        for (var i = 0; i < this.numServers; i++) {
            let app = this.apps[Math.floor(Math.random() * this.apps.length)];
            var server = this.ServersSrvc.create({
                name: this.serverPrefix + " #" + (i + 1),
                appID: app.appID,
                baseURL: "",
                keyType: "",
                publicKey: "",
                permissions: ""
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
        'AppSrvc',
        'ServerSrvc'
    ];
    constructor(
        private $mdDialog: ng.material.IDialogService,
        private AppsSrvc: AppSrvc,
        private ServersSrvc: ServerSrvc,
    ) {
        this.AppsSrvc.query().then((apps) => {
            this.apps = apps;
        })
    }

}