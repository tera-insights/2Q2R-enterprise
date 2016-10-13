/// <reference path="../../typings/index.d.ts" />

/**
 * This service manages admins (if the current admin has permissions to do so) 
 * 
 * @export
 * @class Admins
 */
export class Admins {



    static $inject = ['$http'];
    constructor(private $http: ng.IHttpService) {

    }
}