/// <reference path="../../typings/index.d.ts" />

module admin {

    /**
     * This service manages the applications and the application servers 
     * 
     * @export
     * @class Apps
     */
    export class Apps {
        

        static $inject = ['$http'];
        constructor(private $http: ng.IHttpService) {

        }
    }
}