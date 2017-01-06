import { Apps, IAppItem, IAppResource } from '../../services/Apps';
import { createAuthenticator } from 'p256-auth';
import 'file-saver';

interface NewAdminRequest {
    name: string;
    email: string;
    adminFor: string; // appID
    iv: string;
    salt: string;
    publicKey: string;
}

/**
 * Controller for the registration modal. 
 * 
 * @export
 * @class RegisterCtrl
 */
export class RegisterCtrl {
    private availablePermissions: string[] = [
        'Register',
        'Authenticate',
        'Analytics'
    ];

    private registration: NewAdminRequest = {
        name: '',
        email: '',
        adminFor: undefined,
        iv: undefined,
        salt: undefined,
        publicKey: undefined
    };

    /**
     * Finishes off the registration by generating and saving
     * a key pair, as well as the completed registration request
     * for approval by an admin. Closes the modal.
     */
    accept() {
        let authenticator = createAuthenticator();
        authenticator.generateKeyPair();
        let extKeyPair = authenticator.exportKey(new Uint8Array(16));

        let file = new Blob([this.registration], { type: 'application/json;charset=utf-8' });
        saveAs(file, 'NewRegistration.json');
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
        'Apps'
    ];
    constructor(
        private $mdDialog: ng.material.IDialogService
    ) {
    }
}