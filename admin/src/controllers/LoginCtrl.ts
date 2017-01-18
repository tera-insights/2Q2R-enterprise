import { AuthSrvc } from '../services/AuthSrvc';
import { RegisterCtrl } from './modals/RegisterCtrl'

export class LoginCtrl {

    login(userid: string, password: string) {

    }

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
    ) {

    }

}