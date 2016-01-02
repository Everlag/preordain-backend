/*

Returns the presence of a specified event.

Takes $1 as an eventid.

*/

select count(name) > 0 from mtgtop8.events where eventid=$1;