/// <reference path="../../../typings/index.d.ts" />
/// <reference path="../../services/Servers.ts" />

module admin {

    /**
     * Controller of add server modal. 
     * 
     * @export
     * @class AddServerCtrl
     */
    export class AddServerCtrl {
        private server: IServerItem;
        private appID: string; // selected appID

        /**
         * Accept function. Closes modal
         */
        accept() {
            // no need to pass the semester since calee has it.
            this.$mdDialog.hide(this.server);
        }

        /**
         * Cancel all the actions
         */
        cancel() {
            this.$mdDialog.cancel();
        }

        static $inject = [
            '$mdDialog',
            'Servers'
        ];
        constructor(
            private $mdDialog: ng.material.IDialogService,
            ServersSrvc: Servers
        ) {
             var Server = ServersSrvc.resource;
             this.server = new Server({
                 serverName: "",
                 appID: "",
                 baseURL: "",
                 keyType: "",
                 publicKey: "",
                 permissions: ""
             }); 
        }
    }

}