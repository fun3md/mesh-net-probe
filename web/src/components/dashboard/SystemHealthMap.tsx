import React from 'react';

const SystemHealthMap: React.FC = () => {
    return (
        <div className="rounded-xl border border-border-light dark:border-border-dark bg-panel-light dark:bg-panel-dark p-5 flex flex-col">
            <h3 className="text-lg font-semibold text-text-light-primary dark:text-text-dark-primary mb-4">System Health Map</h3>
            <div
                className="flex-1 w-full bg-center bg-no-repeat aspect-video bg-cover rounded-lg object-cover"
                style={{backgroundImage: 'url("https://lh3.googleusercontent.com/aida-public/AB6AXuD7CSoF6HdGBdctlNznn6LL82OTBKJ6YDo9npX1FeND1Ohkx0CAN2MZC5qE5xUaUNQzamNF1qK9gKmryZEymUYSC4NmCXPvapd4J9zsm776qvHMuxv6xDyqDcD1Ch4KunKwOLghzWYKjSYkofUYf03a3Px0nyG3ENXZgcObcpotn0hAd-y6bTnTNpcMJrURNiHi3BMFsts6QsRUqzGnRjQyklGIWHS0WF-z-R7-ihI_NN-rcOgPBdURbXqh64hjktfCqQnGA8Lm4yo")'}}
                data-alt="World map with color-coded pins indicating probe locations and their health status"
            ></div>
        </div>
    );
};

export default SystemHealthMap;
