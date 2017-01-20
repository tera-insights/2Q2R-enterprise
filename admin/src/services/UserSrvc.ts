import 'angular-resource';

/**************
 * Interfaces *
 **************/



/**
 * This service manages 2FA users across all
 * the application servers.
 * 
 * @author Sam Claus
 * @version 1/18/17
 * @copyright Tera Insights, LLC
 */
export class UserSrvc {

    private resource: any = this.$resource('', {}, {
        
    });

    static $inject = [
        '$resource',
        '$q'
    ];

    constructor(
        private $resource: ng.resource.IResourceService,
        private $q: ng.IQService
    ) {}

}