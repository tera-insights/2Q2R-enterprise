import { UserSrvc } from "../services/UserSrvc";

export class UsersCtrl {

    static $inject = [
        'Users'
    ];
    constructor(
        UsersSrvc: UserSrvc
    ) {
       
    }

}