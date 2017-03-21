import { UserSrvc } from "../services/UserSrvc";
import 'angular-resource';

export class UsersCtrl {

    static $inject = [
        'Users'
    ];
    constructor(
        UsersSrvc: UserSrvc
    ) {
       
    }

}