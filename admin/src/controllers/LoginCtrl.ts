import { AuthSrvc } from '../services/AuthSrvc';
import { RegisterCtrl } from './modals/RegisterCtrl';
import { ExternalKeyPair } from 'p256-auth';

export class LoginCtrl {

    private signingKey: ExternalKeyPair;
    private password: Uint8Array = new Uint8Array(50);

    /**
     * Upload the admin's signing key for primary authentication.
     * @param {File} file The signing key file (JSON).
     */
    uploadSigningKey(file: File) {
        let fileReader = new FileReader();
        fileReader.onload = (event) => {
            this.signingKey = JSON.parse((event.target as any).result);
        };
        fileReader.readAsText(file);
    }

    // probably merge with upload signed key, just need to test animations
    submitKey() {
        $(".nokeyreg__add-key").toggleClass("expanded");
    }

    /**
     * Makes a call to the authentication service with the user's input
     * credentials, resulting in the creation of first-factor headers if
     * a valid key file is uploaded and the password is correct.
     * @param {string} userID The input username.
     */
    login(userID: string) {
        this.AuthSrvc.prepareFirstFactor(this.signingKey, userID, this.password);
    }

    /**
     * Opens a prompt for the first half of a new admin registration.
     */
    register() {
        this.$mdDialog.show({
            controller: RegisterCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/Register.html',
            clickOutsideToClose: true
        }).then(() => {
            this.$mdToast.showSimple('Registration request saved. Email it to your superadmin to get approved.');
        }, () => {
            this.$mdToast.showSimple('Registration canceled.');
        });
    }


    static $inject = [
        'AuthSrvc',
        '$mdDialog',
        '$mdToast'
    ];

    constructor(
        private AuthSrvc: AuthSrvc,
        private $mdDialog: angular.material.IDialogService,
        private $mdToast: angular.material.IToastService
    ) {}

}