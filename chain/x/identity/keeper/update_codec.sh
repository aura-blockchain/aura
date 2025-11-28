#!/bin/bash

# Update auth.go
sed -i 's/json\.Marshal(\([^)]*\))/k.cdc.Marshal(\&\1)/g' auth.go
sed -i 's/json\.Unmarshal(\([^,]*\), \&\([^)]*\))/k.cdc.Unmarshal(\1, \&\2)/g' auth.go

# Update changes.go  
sed -i 's/json\.Marshal(\([^)]*\))/k.cdc.Marshal(\&\1)/g' changes.go
sed -i 's/json\.Unmarshal(\([^,]*\), \&\([^)]*\))/k.cdc.Unmarshal(\1, \&\2)/g' changes.go

# Update sessions.go
sed -i 's/json\.Marshal(\([^)]*\))/k.cdc.Marshal(\&\1)/g' sessions.go
sed -i 's/json\.Unmarshal(\([^,]*\), \&\([^)]*\))/k.cdc.Unmarshal(\1, \&\2)/g' sessions.go

echo "Updated JSON marshal/unmarshal calls to use codec"
