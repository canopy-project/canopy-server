id -g canopy >> /dev/null || addgroup --system canopy
id -u canopy >> /dev/null || adduser --system --ingroup canopy --no-create-home --disabled-password canopy
