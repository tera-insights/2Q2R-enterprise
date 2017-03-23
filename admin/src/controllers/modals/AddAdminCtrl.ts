import { AdminSrvc } from '../../services/AdminSrvc';
import { INewAdminRequest } from '../../services/AdminSrvc';
import { ExternalKeyPair, createAuthenticator } from 'p256-auth';

/**
 * Admin registration modal controller.
 * 
 * @author Sam Claus
 * @version 1/12/17
 * @copyright Tera Insights, LLC
 */
export class AddAdminCtrl {
    private registration: INewAdminRequest;
    private adminKey: ExternalKeyPair

    private availablePermissions: string[] = [
        'Register',
        'Authenticate',
        'Analytics'
    ];

    /**
     * Upload the registering admin's registration file.
     * @param {File} file The registration file (JSON).
     */
    uploadRegistration(file: File) {
        let fileReader = new FileReader();
        fileReader.onload = (event) => {
            this.registration = JSON.parse((event.target as any).result);
        };
        fileReader.readAsText(file);
    }

    /**
     * The referring admin must upload their key so it
     * can be used to sign the registering admin's public
     * key.
     * @param {File} file The signing key file (JSON).
     */
    uploadKey(file: File) {

    }

    /**
     * Send the new admin registration to the server. The modal
     * then either returns void, or an error if something went
     * wrong with the registration.
     */
    accept() {
        this.AdminSrvc.create(this.registration).then(reply => {
            this.$mdDialog.hide();
        }).catch(error => {
            this.$mdDialog.hide(error);
        });
    }

    /**
     * Cancel the registration.
     */
    cancel() {
        this.$mdDialog.cancel();
    }

    static $inject = [
        '$mdDialog',
        'AdminSrvc'
    ];

    constructor(
        private $mdDialog: ng.material.IDialogService,
        private AdminSrvc: AdminSrvc
    ) {
    }
}