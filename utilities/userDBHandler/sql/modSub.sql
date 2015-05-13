/*

UPSERTs into users.subs using the mod_sub function.

Could potentially loop forever but good god does postgres need this natively.

mod_sub has the format
	mod_sub('everlag', 'Sensei\'s Top', now,
	'someCustomerToken', 'someSubToken');

Takes:
	name - string, user that owns it
	plan - possibleSub, the plan name the user desires
	time - timestamp, the latest change
	customerID - string, the customers id as provided by stripe
	subID - string, the sub id as provided by stripe
*/

SELECT mod_sub($1, $2, $3, $4, $5);