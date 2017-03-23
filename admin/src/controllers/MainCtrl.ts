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
                state: "main.reports",
                name: "Reports",
                icon: "mdi mdi-clipboard-text"
            },
            {
                state: "main.servers",
                name: "Servers",
                icon: "mdi mdi-server"
            },
            {
                state: "main.admin",
                name: "Administrators",
                icon: "mdi mdi-certificate"
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
    // is then set based on clicked menuitem , defaults to dash
    private activeMenu: string = "Dashboard";

    select(route: string, name: string, menu: string) {
        this.sName = name;
            
        // go to the selected state    
        this.$state.go(route);

        // set the active item
        this.activeMenu = menu;
    }

    generate() {
        this.$mdDialog.show({
            controller: GenerateCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/Generate.html',
            clickOutsideToClose: true
        }).then(() => {
            // Yay. Generation succesful
        }, () => {
            // TODO: Add notifications
        });
    }

    // used in below function
    private randomNumber: number = Math.floor(Math.random() * 6) + 1;
    // displayed on the copyright
    private randomHtmlChar: string;

    // generates a random number and puts an html symbol according to that on the copyright
    generateRandomHtmlChar() {
        switch (this.randomNumber) {
            case 0:
                this.randomHtmlChar = "&#9733;"; // star
                break;
            case 1:
                this.randomHtmlChar = "&#9786;"; // smiley face
                break;
            case 2:
                this.randomHtmlChar = "&hearts;"; // heart
                break;
            case 3:
                this.randomHtmlChar = "&#9834;"; // note
                break;
            case 4:
                this.randomHtmlChar = "&#36;"; // dollar sign
                break;
            case 5:
                this.randomHtmlChar = "&infin;"; // infinity sign
                break;
            case 6:
                this.randomHtmlChar = "&spades;"; // spades
        }

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
        // set the default state (dashboard)
        this.select("main.dashboard", "Dashboard", "dashboard");

        // generate copyright special character
        this.generateRandomHtmlChar();
    }
}