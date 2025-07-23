CREATE TABLE IF NOT EXISTS "media" (
"id" integer	 primary key,
"media_type" text not null	,
"title"	text not null,
"year"	integer not null,
"overview" text	,
"director" text	,
"poster"	binary,
"path" text not null	
);

