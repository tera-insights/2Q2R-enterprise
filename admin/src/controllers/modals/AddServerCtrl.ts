import { AppSrvc, IAppInfo } from '../../services/AppSrvc';
import { ServerSrvc, IServerInfo } from '../../services/ServerSrvc';

/**
 * Controller of add server modal. 
 * 
 * @export
 * @class AddServerCtrl
 */
export class AddServerCtrl {
    private server: IServerInfo = {
                serverID: "",
                appID: "",
                baseURL: "",
                keyType: "P-256",
                publicKey: "",
                permissions: ""
            };
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
        'ServerSrvc'
    ];
    constructor(
        private $mdDialog: ng.material.IDialogService,
        private ServersSrvc: ServerSrvc
    ) {

    }
}