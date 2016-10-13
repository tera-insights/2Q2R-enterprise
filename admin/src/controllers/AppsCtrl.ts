/// <reference path="../../typings/index.d.ts" />

import { Apps, IAppItem, IAppResource } from '../services/Apps';
import { AddAppCtrl } from '../controllers/modals/AddAppCtrl';

/**
 * Controller for listing and manipulating apps the apps 
 * 
 * @export
 * @class AppsCtrl
 */
export class AppsCtrl {
    private App: IAppResource;

    private apps: IAppItem[] = [];

    // Triggered by the FAB
    newApp() {
        this.$mdDialog.show({
            controller: AddAppCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/AddApp.html',
            clickOutsideToClose: true
        }).then((app: IAppItem) => {
            app.$save().then((newApp: IAppItem) => {
                this.apps.push(newApp);
            });
        }, () => {
            // TODO: Add nofitifications
        });
    }

    updateApp(app: IAppItem) {
        app.$update();
    }

    removeApp(app) {
        var $index = -1;
        this.apps.forEach((t, i, a) => {
            if (t.appID == app.appID)
                $index = i;
        });
        // Take out current element from todos array
        if ($index >= 0) {
            app.$delete();
            this.apps.splice($index, 1);
        }
    }

    static $inject = [
        '$mdDialog',
        'Apps'
    ];

    constructor(
        private $mdDialog: ng.material.IDialogService,
        AppsSrvc: Apps
    ) {
        this.App = AppsSrvc.resource;

        this.apps = this.App.query();
    }

}