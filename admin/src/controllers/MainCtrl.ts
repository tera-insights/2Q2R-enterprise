/// <reference path="../../typings/index.d.ts" />

module admin {
    export class MainCtrl {
        private sName: string = "";

        /**
         * Select a sub-view 
         * 
         * @param {string} route This is the route to switch to
         * @param {string} name This is the name of the state to display
         */
        select(route: string, name: string){
            this.sName = name; 
            this.$state.go(route);
        }

        toggleLeft(){
            this.$mdSidenav('left').toggle();
        }

        static $inject = [
            '$mdSidenav',
            '$state'
            ];
        constructor(
            private $mdSidenav: ng.material.ISidenavService,
            private $state: angular.ui.IStateService
        ) {
            this.toggleLeft();
        }
    }
}