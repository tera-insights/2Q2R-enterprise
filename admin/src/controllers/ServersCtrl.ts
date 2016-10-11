/// <reference path="../../typings/index.d.ts" />
/// <reference path="../services/Servers.ts" />
/// <reference path="../controllers/modals/AddServerCtrl.ts" />

module admin {

    export class ServersCtrl {

        private Server: IServerResource;
        private servers: IServerItem[] = [];

        newServer() {
            this.$mdDialog.show({
                controller: AddServerCtrl,
                controllerAs: 'cMod',
                templateUrl: 'views/modals/AddServer.html',
                clickOutsideToClose: true
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

        static $inject = [
            '$mdDialog',
            'Servers'
        ];
        constructor(
            private $mdDialog: ng.material.IDialogService,
            ServersSrvc: Servers
        ) {
            this.Server = ServersSrvc.resource;
            this.servers = this.Server.query();
        }

    }

}