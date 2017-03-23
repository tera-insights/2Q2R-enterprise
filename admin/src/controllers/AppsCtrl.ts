import { AppSrvc, IAppInfo } from '../services/AppSrvc';
import { AddAppCtrl } from '../controllers/modals/AddAppCtrl';

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

    // adds a new application
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

    // triggers a sheet to move into view on a apps__tile
    settingsTrigger() {
        // add a class that just moves it up
        $(".settings__sheet").addClass("settings__sheet--exists").delay(100).queue(function(next){
            $(this).addClass("settings__sheet--animated");
            next();
        });

    }

    // when the sheet is up, click a button to move it back down
    settingsRetract() {
        // now take it away
        $(".settings__sheet").addClass("settings__sheet--deanimated").delay(450).queue(function(next){
            $(this).removeClass("settings__sheet--animated");
            next();
            $(this).removeClass("settings__sheet--deanimated");
            next();
            $(this).removeClass("settings__sheet--exists");
            next();
        });
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