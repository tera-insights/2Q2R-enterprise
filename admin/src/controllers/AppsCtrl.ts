/// <reference path="../../typings/index.d.ts" />
/// <reference path="../services/Apps.ts" />

module admin {
    /**
     * Controller for listing and manipulating apps the apps 
     * 
     * @export
     * @class AppsCtrl
     */
    export class AppsCtrl {
        private App: IAppResource;

        private apps: IAppItem[] = [];

        addApp(appName: string, id: string) {
            var app = new this.App({
                appName: appName,
                id: id
            });
            app.$save();
            this.apps.push(app);
        }

        updateApp(app: IAppItem) {
            app.$update();
        }

        removeApp(app) {
            var $index = -1;
            this.apps.forEach((t, i, a) => {
                if (t.id == app.id)
                    $index = i;
            });
            // Take out current element from todos array
            if ($index >= 0) {
                app.$delete();
                this.apps.splice($index, 1);
            }
        }

        static $inject = [
            'Apps'
        ];

        constructor(
            AppsSrvc: Apps
        ) {
            this.App = AppsSrvc.resource;

            this.apps = this.App.query();
        }

    }

}