
/// <reference path="../../../typings/index.d.ts" />
/// <reference path="../../services/Apps.ts" />

module admin {

    /**
     * Controller of add app modal. 
     * 
     * @export
     * @class AddAppCtrl
     */
    export class AddAppCtrl {
        private app: IAppItem;

        /**
         * Accept function. Closes modal
         */
        accept() {
            // no need to pass the semester since calee has it.
            this.$mdDialog.hide(this.app);
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
            private $mdDialog: ng.material.IDialogService,
            AppsSrvc: Apps
        ) {
             var App = AppsSrvc.resource;
             this.app = new App({
                 appName: ""
             }); 
        }
    }

}