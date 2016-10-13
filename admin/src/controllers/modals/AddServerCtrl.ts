/// <reference path="../../../typings/index.d.ts" />

import { Apps, IAppItem, IAppResource } from '../../services/Apps';
import { Servers, IServerItem, IServerResource } from '../../services/Servers';

/**
 * Controller of add server modal. 
 * 
 * @export
 * @class AddServerCtrl
 */
export class AddServerCtrl {
    private server: IServerItem;
    private availablePermissions: string[] = [
        'Register',
        'Authenticate',
        'Analytics'
    ];

    private selectedPermissions: string[] = [];

    /**
     * Accept function. Closes modal
     */
    accept() {
        // no need to pass the semester since calee has it.
        this.server.permissions = this.selectedPermissions.join(',');
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
        'Servers',
        'apps'
    ];
    constructor(
        private $mdDialog: ng.material.IDialogService,
        ServersSrvc: Servers,
        private apps: IAppItem[]
    ) {
            var Server = ServersSrvc.resource;
            this.server = new Server({
                serverName: "",
                appID: "",
                baseURL: "",
                keyType: "P-256",
                publicKey: "",
                permissions: ""
            });
    }
}