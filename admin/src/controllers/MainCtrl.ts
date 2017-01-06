import { Auth } from '../services/Auth';
import { GenerateCtrl } from './modals/GenerateCtrl';

interface IMenuItem {
    state: string; // State corresponding to the menu item
    name: string; // This is the displayable name of the menu
    icon: string; // icon name. Assumed to be in /img/icons
}

export class MainCtrl {
    private sName: string = "";

    private menuGroups: IMenuItem[][] = [
        [
            {
                state: "main.dashboard",
                name: "Dashboard",
                icon: "dashboard.svg"
            },
            {
                state: "main.admin",
                name: "Administrators",
                icon: "admin.svg"
            }
        ], [
            {
                state: "main.apps",
                name: "Applications",
                icon: "application.svg"
            },
            {
                state: "main.servers",
                name: "Servers",
                icon: "server.svg"
            },
            {
                state: "main.users",
                name: "Users",
                icon: "user.svg"
            },
            {
                state: "main.2FA",
                name: "2FA Devices",
                icon: "2FA.svg"
            },
            {
                state: "main.reports",
                name: "Reports",
                icon: "reports.svg"
            },
            {
                state: "main.settings",
                name: "Settings",
                icon: "settings.svg"
            }],
        [
            {
                state: "login",
                name: "Logout",
                icon: "logout.svg"
            }]
    ];

    /**
     * Select a sub-view 
     * 
     * @param {string} route This is the route to switch to
     * @param {string} name This is the name of the state to display
     */
    select(route: string, name: string) {
        this.sName = name;
        this.$state.go(route);

        // and toggle the sidenav
        this.$mdSidenav('left').toggle();
    }

    toggleLeft() {
        this.$mdSidenav('left').toggle();
    }

    generate() {
        this.$mdDialog.show({
            controller: GenerateCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/Generate.html',
            clickOutsideToClose: true
        }).then(() => {
            // Yey. Generation succesful
        }, () => {
            // TODO: Add notifications
        });
    }

    static $inject = [
        '$mdSidenav',
        '$state',
        '$mdDialog',
        'Auth'
    ];
    constructor(
        private $mdSidenav: ng.material.ISidenavService,
        private $state: angular.ui.IStateService,
        private $mdDialog: ng.material.IDialogService,
        private Auth: Auth
    ) {
        this.toggleLeft();
        this.select("main.dashboard", "Dashboard");
    }
}