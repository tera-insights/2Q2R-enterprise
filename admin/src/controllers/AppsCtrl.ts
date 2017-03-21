import { AppSrvc, IAppInfo } from '../services/AppSrvc';
import { AddAppCtrl } from '../controllers/modals/AddAppCtrl';
import 'angular-resource';
import 'angular-material';

/**
 * Controller for listing and manipulating apps the apps 
 * 
 * @export
 * @class AppsCtrl
 */
export class AppsCtrl {
    private apps: IAppInfo[] = [];

    // selected items
    private selected: IAppInfo[] = [];

    // angular-material-data-table options
    private options = {
        rowSelect: true,
        autoSelect: true,
        multiSelect: true    }

    private tableQuery = {
        limit: 14,
        page: 1,
        order: "name"
    }

    // Triggered by the FAB
    newApp() {
        this.$mdDialog.show({
            controller: AddAppCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/AddApp.html',
            clickOutsideToClose: true
        }).then((app: IAppInfo) => {
            this.AppSrvc.create({
                appName: app.appName
            }).then( (a) => {
                this.apps.push(a);
            });
        }, () => {
            // TODO: Add nofitifications
        });
    }

    updateApp(app: IAppInfo) {
        this.AppSrvc.update(app);
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
        'AppSrvc'
    ];

    constructor(
        private $mdDialog: ng.material.IDialogService,
        private AppSrvc: AppSrvc
    ) {
       this.AppSrvc.query().then( (apps) => {
           this.apps = apps;
       });
    }

}