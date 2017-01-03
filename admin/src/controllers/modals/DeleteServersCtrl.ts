/**
 * Controller of delete servers modal. 
 * 
 * @export
 * @class DeleteServersCtrl
 */
export class DeleteServersCtrl {

    /**
     * Accept function. Closes modal
     */
    accept() {
        this.$mdDialog.hide();
    }

    /**
     * Cancel all the actions
     */
    cancel() {
        this.$mdDialog.cancel();
    }

    static $inject = [
        '$mdDialog'
    ];
    constructor(
        private $mdDialog: ng.material.IDialogService
    ) {}
}