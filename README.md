# catlitter

a quick little validator for places that generate short urls, such as imgur, catbox, etc.

Gnerate your (purrr)mutations (get it now) using crunch, I like to pipe that into split -l 100000 giving me many 100,000 line files in a directory

./catlitter -d /path/to/dir/of/lists -url https://base.url/ -ext .jpg (or any other file, or -no-ext for naked files)

wait 37 years (approx) to cover all possible combinations of A-Za-z0-9 at 6 chars

profit.

[Geonode](https://app.geonode.com/invite/24053) is now implemented, you will need to rename proxy_config.json.example to proxy_config.json and fill in your details from the geonode dash, 

this will help with rate limits as you can rotate through millions of IPs automatically

