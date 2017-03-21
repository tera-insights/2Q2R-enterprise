import { AuthSrvc } from '../services/AuthSrvc';
import { GenerateCtrl } from './modals/GenerateCtrl';

interface IMenuItem {
    state: string; // State corresponding to the menu item
    name: string; // This is the displayable name of the menu
    icon: string; // icon name. Assumed to be in /img/icons
}

export class MainCtrl {
    private sName: string = "";

    // controls what the main left menu displays, in main.html
    private menuGroups: IMenuItem[][] = [
        [
            {
                // what state it redirects to
                state: "main.dashboard",
                // The text that shows up next to the icon
                name: "Dashboard",
                // the icon, lights up with a color if clicked
                icon: "mdi mdi-view-dashboard"
            },
            {
                state: "main.admin",
                name: "Administrators",
                icon: "mdi mdi-certificate"
            },
            {
                state: "main.apps",
                name: "Applications",
                icon: "mdi mdi-cloud"
            },
            {
                state: "main.servers",
                name: "Servers",
                icon: "mdi mdi-server"
            },
            {
                state: "main.users",
                name: "Users",
                icon: "mdi mdi-account"
            },
            {
                state: "main.2FA",
                name: "2FA Devices",
                icon: "mdi mdi-cellphone-link"
            },
            {
                state: "main.reports",
                name: "Reports",
                icon: "mdi mdi-clipboard-text"
            },
            {
                state: "main.settings",
                name: "Settings",
                icon: "mdi mdi-settings"
            },
            {
                state: "login",
                name: "Logout",
                icon: "mdi mdi-logout-variant"
            }
        ]    
    ];

    /**
     * Select a sub-view 
     * 
     * @param {string} route This is the route to switch to
     * @param {string} name This is the name of the state to display
    */

    // a variable that manages the active navigator
    // is then set based on clicked menuitem 
    private activeMenu: string;

    select(route: string, name: string, menu: string) {
        this.sName = name;
            
        // go to the selected state    
        this.$state.go(route);

        // set the active item
        this.activeMenu = menu;

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
        'AuthSrvc'
    ];
    constructor(
        private $mdSidenav: ng.material.ISidenavService,
        private $state: angular.ui.IStateService,
        private $mdDialog: ng.material.IDialogService,
        private AuthSrvc: AuthSrvc
    ) {
        this.toggleLeft();
        this.select("main.dashboard", "Dashboard", "???");
    }
}