import { AppSrvc, IAppInfo } from '../../services/AppSrvc';

/**
 * Controller of add app modal. 
 * 
 * @export
 * @class AddAppCtrl
 */
export class AddAppCtrl {
    private appName: string = "";

    /**
     * Accept function. Closes modal
     */
    accept() {
        // no need to pass the semester since calee has it.
        this.$mdDialog.hide(this.appName);
    }

    /**
     * Cancel all the actions
     */
    cancel() {
        this.$mdDialog.cancel();
    }

    static $inject = [
        '$mdDialog',
        'AppSrvc'
    ];
    constructor(
        private $mdDialog: ng.material.IDialogService,
        private AppsSrvc: AppSrvc
    ) {        
    }
}