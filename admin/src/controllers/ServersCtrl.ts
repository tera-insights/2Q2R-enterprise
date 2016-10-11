/// <reference path="../../typings/index.d.ts" />
/// <reference path="../services/Servers.ts" />
/// <reference path="../services/Apps.ts" />
/// <reference path="../controllers/modals/AddServerCtrl.ts" />

module admin {

    export class ServersCtrl {
        private Server: IServerResource;
        private App: IAppResource;

        private apps: IAppItem[] = [];
        private servers: IServerItem[] = [];

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
            this.servers = this.Server.query();
            this.apps = this.App.query();
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

}