/// <reference path="../../typings/index.d.ts" />

module admin {

    /**
     * Service responsible for logging in and out and maintaining sessions. 
     * 
     * @export
     * @class Auth
     */
    export class Auth {
        private userid: string; // our user id
        private userName: string; // our name

        // TODO: add persmissions

        private loggedIn: boolean = false; // are we logged in?

        

        static $inject = ['$http'];
        constructor(private $http: ng.IHttpService) {
            this.userid = "alin@terainsighs.com";
            this.userName = "Alin Dobra";
        }
    }

}
