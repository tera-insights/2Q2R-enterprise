import { INewAdminRequest } from '../../interfaces/rest';
import { createAuthenticator } from 'p256-auth';
import FileSaver = require('file-saver');

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

    private registration: INewAdminRequest = {
        name: '',
        email: '',
        adminFor: '',
        iv: undefined,
        salt: undefined,
        publicKey: undefined
    };

    private password: Uint8Array = new Uint8Array(50);

    /**
     * Finishes off the registration by generating and saving
     * a key pair, as well as the completed registration request
     * for approval by an admin. Closes the modal.
     */
    accept() {
        let authenticator = createAuthenticator();
        
        authenticator.generateKeyPair().then(() => {
            authenticator.exportKey(this.password).then( (extKey) => 
                authenticator.getPublic().then((pubKey) => {
                this.registration.iv = extKey.iv;
                this.registration.salt = extKey.salt;
                this.registration.publicKey = pubKey;

                let keyFile = new Blob([JSON.stringify(extKey, null, 2)], { type: 'text/json;charset=utf-8' });
                let regFile = new Blob([JSON.stringify(this.registration, null, 2)], { type: 'text/json;charset=utf-8' });

                FileSaver.saveAs(keyFile, 'Key.1fa');
                FileSaver.saveAs(regFile, this.registration.name.replace(' ', '_') + '.arr');
                this.$mdDialog.hide();
            }))
        });
    }

    /**
     * Cancel all the actions
     */
    cancel() {
        this.$mdDialog.cancel();
    }

    static $inject = [
        '$mdDialog',
        '$q'
    ];
    constructor(
        private $mdDialog: ng.material.IDialogService,
        private $q: ng.IQService
    ) {
        console.log(this.password[0]);
    }
}